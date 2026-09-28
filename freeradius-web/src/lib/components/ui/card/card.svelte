<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<div class={cn('rounded-xl border bg-card text-card-foreground shadow', klass)} {...$$restProps}>
	<slot />
</div>
